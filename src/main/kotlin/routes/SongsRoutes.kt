package com.witelokk.routes

import com.witelokk.models.FailureResponse
import com.witelokk.tables.Artists
import com.witelokk.tables.Favorites
import com.witelokk.tables.SongArtists
import com.witelokk.tables.Songs
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import models.ShortArtist
import models.Song
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.UUID

fun Route.songsRoutes() {
    authenticate("auth-jwt") {
        route("/songs") {
            get("/{id}") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val songId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("artist_not_found", "Artist not found")
                    )
                }

                val song = transaction {
                    getSongWithArtistsAndFavorite(songId, userId)
                }

                if (song == null) {
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("song_not_found", "Song not found")
                    )
                }

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
                ShortArtist(
                    id = artistRow[Artists.id].toString(),
                    name = artistRow[Artists.name],
                    avatarUrl = artistRow[Artists.avatarUrl]
                )
            }

        val isFavorite = userId?.let {
            Favorites.select { (Favorites.songId eq songId) and (Favorites.userId eq it) }
                .count() > 0
        } ?: false

        Song(
            id = songRow[Songs.id].toString(),
            name = songRow[Songs.name],
            coverUrl = songRow[Songs.coverUrl],
            isFavorite = isFavorite,
            durationSeconds = songRow[Songs.duration],
            artists = artists,
            streamUrl = songRow[Songs.streamUrl]
        )
    }
}