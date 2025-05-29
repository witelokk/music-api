package com.witelokk.music.routes

import com.witelokk.music.models.FailureResponse
import io.github.smiley4.ktoropenapi.route
import io.github.smiley4.ktoropenapi.get
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import com.witelokk.music.models.Artist
import com.witelokk.music.tables.*
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.*

fun Route.artistsRoutes() {
    authenticate("auth-jwt") {
        route("/artists", {
            tags = listOf("artists")
        }) {
            get("/{id}", {
                description = "Get artist by ID"
                request {
                    pathParameter<String>("id") {
                        description = "Artist ID"
                    }
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<Artist>()
                    }
                    HttpStatusCode.NotFound to {
                        description = "Artist not found"
                        body<FailureResponse>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val artistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("artist_not_found", "Artist not found")
                    )
                }

                val artist = transaction {
                    getArtistWithFollowingAndPopularSongAndRecentReleases(artistId, userId)
                } ?: return@get call.respond(
                    HttpStatusCode.NotFound, FailureResponse("artist_not_found", "Artist not found")
                )

                return@get call.respond(artist)
            }
        }
    }
}

private fun getArtistWithFollowingAndPopularSongAndRecentReleases(artistId: UUID, userId: UUID): Artist? {
    val followingAlias = exists(
        Followers.select {
            (Followers.artistId eq artistId) and (Followers.userId eq userId)
        }
    ).alias("following")

    val popularSongs =
        SongArtists.leftJoin(Songs).select { SongArtists.artistId eq artistId }.limit(5)
            .sortedBy { Songs.streamsCount }
            .map {
                getSongWithArtistsAndFavorite(it[SongArtists.songId], userId)
            }.filter { it != null }.map { it!! }
    val releases =
        ReleaseArtists.leftJoin(Releases)
            .select { ReleaseArtists.artistId eq artistId }
            .orderBy(Releases.releasedAt, org.jetbrains.exposed.sql.SortOrder.DESC)
            .map {
                getReleaseWithArtist(userId, it[ReleaseArtists.releaseId])
            }.filter { it != null }.map { it!! }

    val q = Artists.leftJoin(Followers).slice(
//        Artists.id, Artists.name, Artists.coverUrl, Artists.avatarUrl, Followers.userId.count(), followingAlias
        Artists.fields + Followers.userId.count() + followingAlias
    ).select { Artists.id eq artistId }
        .groupBy(*Artists.fields.toTypedArray()).map {
            Artist(
                id = it[Artists.id],
                name = it[Artists.name],
                avatarUrl = it[Artists.avatarUrl],
                coverUrl = it[Artists.coverUrl],
                followers = it[Followers.userId.count()].toInt() ?: 0,
                following = it[followingAlias],
                popularSongs = com.witelokk.music.models.Songs(
                    count = popularSongs.size,
                    songs = popularSongs,
                ),
                releases = com.witelokk.music.models.Releases(
                    count = releases.size,
                    releases = releases,
                )
            )
        }

    return q.singleOrNull()
}