package com.witelokk.music.routes

import com.witelokk.music.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.music.PG_UNIQUE_VIOLATION
import com.witelokk.music.models.AddSongToPlaylistRequest
import com.witelokk.music.models.FailureResponse
import com.witelokk.music.models.RemoveSongFromPlaylistRequest
import com.witelokk.music.models.Songs
import com.witelokk.music.tables.Playlists
import com.witelokk.music.tables.PlaylistSongs
import io.github.smiley4.ktorswaggerui.dsl.routing.delete
import io.github.smiley4.ktorswaggerui.dsl.routing.get
import io.github.smiley4.ktorswaggerui.dsl.routing.post
import io.github.smiley4.ktorswaggerui.dsl.routing.route
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
        route("/playlists/{id}/songs", {
            tags = listOf("playlists")
        }) {
            get("/", {
                description = "Get playlist songs"
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<Songs>()
                    }
                    HttpStatusCode.BadRequest to {
                        description = "Bad request"
                        body<FailureResponse>()
                    }
                }
            }) {
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

            post("/", {
                description = "Add song to playlist"
                request {
                    body<AddSongToPlaylistRequest>()
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                    }
                    HttpStatusCode.BadRequest to {
                        description = "Bad request"
                        body<FailureResponse>()
                    }
                }
            }) {
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
                    addSongToPlaylist(playlistId, request.songId)
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

            delete("/", {
                description = "Remove song from playlist"
                request {
                    body<RemoveSongFromPlaylistRequest>()
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                    }
                    HttpStatusCode.BadRequest to {
                        description = "Bad request"
                        body<FailureResponse>()
                    }
                }
            }) {
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
                    removeSongFromPlaylist(playlistId, request.songId)
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

fun getPlaylistSongs(userId: UUID, playlistId: UUID): Songs {
    val songs = transaction {
        PlaylistSongs
            .select { PlaylistSongs.playlistId eq playlistId }
            .map {
                getSongWithArtistsAndFavorite(it[PlaylistSongs.songId], userId)!!
            }
    }
    return Songs(
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