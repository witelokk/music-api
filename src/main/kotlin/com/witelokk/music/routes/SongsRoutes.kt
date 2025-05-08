package com.witelokk.music.routes

import com.witelokk.music.models.FailureResponse
import com.witelokk.music.tables.Artists
import com.witelokk.music.tables.Favorites
import com.witelokk.music.tables.SongArtists
import com.witelokk.music.tables.Songs
import io.github.smiley4.ktoropenapi.get
import io.github.smiley4.ktoropenapi.route
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import com.witelokk.music.models.AristSummary
import com.witelokk.music.models.Song
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.UUID

fun Route.songsRoutes() {
    authenticate("auth-jwt") {
        route("/songs", {
            tags = listOf("songs")
        }) {
            get("/{id}", {
                description = "Get song by ID"
                request {
                    pathParameter<String>("id") {
                        description = "Song ID"
                    }
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<Song>()
                    }
                    HttpStatusCode.NotFound to {
                        description = "Song not found"
                        body<FailureResponse>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val songId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("artist_not_found", "Artist not found")
                    )
                }

                val song = getSongWithArtistsAndFavorite(songId, userId)
                    ?: return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("song_not_found", "Song not found")
                    )

                call.respond(song);
            }
        }
    }
}

fun getSongWithArtistsAndFavorite(songId: UUID, userId: UUID?): Song? {
    return transaction {
        val songRow = Songs.select { Songs.id eq songId }.singleOrNull() ?: return@transaction null

        val artists = (SongArtists innerJoin Artists)
            .select { SongArtists.songId eq songId }
            .map { artistRow ->
                AristSummary(
                    id = artistRow[Artists.id],
                    name = artistRow[Artists.name],
                    avatarUrl = artistRow[Artists.avatarUrl]
                )
            }

        val isFavorite = userId?.let {
            Favorites.select { (Favorites.songId eq songId) and (Favorites.userId eq it) }
                .count() > 0
        } ?: false

        Song(
            id = songRow[Songs.id],
            name = songRow[Songs.name],
            coverUrl = songRow[Songs.coverUrl],
            isFavorite = isFavorite,
            durationSeconds = songRow[Songs.duration],
            artists = artists,
            streamUrl = songRow[Songs.streamUrl]
        )
    }
}