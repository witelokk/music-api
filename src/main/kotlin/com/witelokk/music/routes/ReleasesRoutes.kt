package com.witelokk.music.routes

import com.witelokk.music.models.FailureResponse
import com.witelokk.music.models.Release
import com.witelokk.music.models.ArtistsSummary
import com.witelokk.music.models.Songs
import io.github.smiley4.ktoropenapi.get
import io.github.smiley4.ktoropenapi.route
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import com.witelokk.music.models.ArtistSummary
import com.witelokk.music.tables.Artists
import com.witelokk.music.tables.ReleaseArtists
import com.witelokk.music.tables.ReleaseSongs
import com.witelokk.music.tables.Releases
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.*

fun Route.releasesRoutes() {
    authenticate("auth-jwt") {
        route("/releases", {
            tags = listOf("releases")
        }) {
            get("/{id}", {
                description = "Get release by ID"
                request {
                    pathParameter<String>("id") {
                        description = "Release ID"
                    }
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<Release>()
                    }
                    HttpStatusCode.NotFound to {
                        description = "Release not found"
                        body<FailureResponse>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val releaseId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("release_not_found", "Release not found")
                    )
                }

                val release = getReleaseWithArtist(userId, releaseId)
                    ?: return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("release_not_found", "Release not found")
                    )

                call.respond(release)
            }
        }
    }
}

fun getReleaseWithArtist(userId: UUID, releaseId: UUID): Release? {
    return transaction {
        val release = Releases.select { Releases.id eq releaseId }.singleOrNull() ?: return@transaction null

        val artists = ReleaseArtists.leftJoin(Artists)
            .select { ReleaseArtists.releaseId eq releaseId }
            .map { artistRow ->
                ArtistSummary(
                    id = artistRow[Artists.id],
                    name = artistRow[Artists.name],
                    avatarUrl = artistRow[Artists.avatarUrl]
                )
            }

        val songs = ReleaseSongs
            .select { ReleaseSongs.releaseId eq releaseId }
            .map { getSongWithArtistsAndFavorite(it[ReleaseSongs.songId], userId)!! }

        Release(
            id = release[Releases.id],
            name = release[Releases.name],
            coverUrl = release[Releases.coverUrl],
            type = release[Releases.type].name,
            releasedAt = release[Releases.releasedAt],
            songs = Songs(
                count = songs.size,
                songs = songs
            ),
            artists = ArtistsSummary(
                count = artists.size,
                artists = artists
            )
        )
    }
}
