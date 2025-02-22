package com.witelokk.routes

import com.witelokk.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.PG_UNIQUE_VIOLATION
import com.witelokk.models.AddSongToPlaylistRequest
import com.witelokk.models.FailureResponse
import com.witelokk.models.RemoveSongFromPlaylistRequest
import com.witelokk.tables.Playlists
import com.witelokk.tables.PlaylistSongs
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.and
import org.jetbrains.exposed.sql.deleteWhere
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import org.joda.time.DateTime
import java.sql.SQLException
import java.util.*

fun Route.playlistSongsRoutes() {
    authenticate("auth-jwt") {
        route("/playlists/{id}/songs") {
            get("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val playlistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@get call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                val playlistUserId = transaction {
                    Playlists.select { Playlists.id eq playlistId }.map { it[Playlists.userId] }.singleOrNull()
                }

                if (playlistUserId != userId) {
                    return@get call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                val playlistSongs = getPlaylistSongs(userId, playlistId)

                call.respond(playlistSongs)
            }

            post("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<AddSongToPlaylistRequest>()

                val playlistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@post call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }


                val playlistUserId = transaction {
                    Playlists.select { Playlists.id eq playlistId }.map { it[Playlists.userId] }.singleOrNull()
                }

                if (playlistUserId != userId) {
                    return@post call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                try {
                    addSongToPlaylist(playlistId, UUID.fromString(request.songId))
                } catch (e: IllegalArgumentException) {
                    return@post call.respond(
                        HttpStatusCode.BadRequest,
                        FailureResponse("song_not_fount", "Song not found")
                    )
                } catch (e: SQLException) {
                    if (e.sqlState == PG_FOREIGN_KEY_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("song_not_fount", "Song not found")
                        )
                    } else if (e.sqlState == PG_UNIQUE_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("already_favorite", "Song is already in playlist")
                        )
                    }
                } catch (e: Exception) {
                    println(e)
                    return@post call.respond(
                        HttpStatusCode.InternalServerError,
                        FailureResponse("internal_error", "Internal error")
                    )
                }

                call.respond(HttpStatusCode.OK)
            }

            delete("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<RemoveSongFromPlaylistRequest>()

                val playlistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@delete call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                val playlistUserId = transaction {
                    Playlists.select { Playlists.id eq playlistId }.map { it[Playlists.userId] }.singleOrNull()
                }

                if (playlistUserId != userId) {
                    return@delete call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                try {
                    removeSongFromPlaylist(playlistId, UUID.fromString(request.songId))
                } catch (e: IllegalArgumentException) {
                    return@delete call.respond(
                        HttpStatusCode.BadRequest,
                        FailureResponse("song_not_fount", "Song not found")
                    )
                } catch (e: Exception) {
                    println(e)
                    return@delete call.respond(
                        HttpStatusCode.InternalServerError,
                        FailureResponse("internal_error", "Internal error")
                    )
                }

                call.respond(HttpStatusCode.OK)
            }
        }
    }
}

fun getPlaylistSongs(userId: UUID, playlistId: UUID): com.witelokk.models.Songs {
    val songs = transaction {
        PlaylistSongs
            .select { PlaylistSongs.playlistId eq playlistId }
            .map {
                getSongWithArtistsAndFavorite(it[PlaylistSongs.songId], userId)!!
            }
    }
    return com.witelokk.models.Songs(
        count = songs.count(),
        songs = songs
    )
}

fun addSongToPlaylist(playlistId: UUID, songId: UUID) {
    transaction {
        PlaylistSongs.insert {
            it[PlaylistSongs.playlistId] = playlistId
            it[PlaylistSongs.songId] = songId
            it[PlaylistSongs.addedAt] = DateTime.now()
        }
    }
}

fun removeSongFromPlaylist(playlistId: UUID, songId: UUID) {
    transaction {
        PlaylistSongs.deleteWhere { (PlaylistSongs.playlistId eq playlistId) and (PlaylistSongs.songId eq songId) }
    }
}