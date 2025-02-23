package com.witelokk.music.routes

import com.witelokk.music.models.*
import com.witelokk.music.tables.Playlists
import com.witelokk.music.tables.PlaylistSongs
import com.witelokk.music.tables.Songs
import io.github.smiley4.ktorswaggerui.dsl.routing.*
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.transactions.transaction
import org.joda.time.DateTime
import java.util.*

fun Route.playlistsRoutes() {
    authenticate("auth-jwt") {
        route("/playlists", {
            tags = listOf("playlists")
        }) {
            get("/{id}", {
                description = "Get playlist by ID"
                request {
                    pathParameter<String>("id") {
                        description = "Playlist ID"
                    }
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<Playlist>()
                    }
                    HttpStatusCode.NotFound to {
                        description = "Playlist not found"
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
                        HttpStatusCode.NotFound, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                val playlist = getPlayListWithSongs(userId, playlistId)
                    ?: return@get call.respond(
                        HttpStatusCode.NotFound,
                        FailureResponse("playlist_not_found", "Playlist not found")
                    )

                call.respond(playlist)
            }

            get("/", {
                description = "Get a list of playlists"
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<ShortPlaylists>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val playlists = getPlaylists(userId)

                call.respond(playlists)
            }

            post("/", {
                description = "Create a new playlist"
                request {
                    body<CreatePlaylistRequest>()
                }
                response {
                    HttpStatusCode.Created to {
                        description = "Success"
                        body<CreatePlaylistResponse>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<CreatePlaylistRequest>()

                val id = UUID.randomUUID()
                createPlaylist(id, userId, request)
                call.respond(HttpStatusCode.Created, CreatePlaylistResponse(id))
            }

            delete("/{id}", {
                description = "Delete a playlist"
                request {
                    pathParameter<String>("id") {
                        description = "Playlist ID"
                    }
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

                val playlistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@delete call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                deletePlaylist(playlistId, userId)
            }

            put("/{id}", {
                description = "Update a playlist"
                request {
                    pathParameter<String>("id") {
                        description = "Playlist ID"
                    }
                    body<UpdatePlaylistRequest>()
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
                val request = call.receive<UpdatePlaylistRequest>()

                val playlistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@put call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                val playlistUserId = transaction {
                    Playlists.select { Playlists.id eq playlistId }.map { it[Playlists.userId] }.singleOrNull()
                }

                if (playlistUserId != userId) {
                    return@put call.respond(
                        HttpStatusCode.BadRequest, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                updatePlaylist(playlistId, request)
            }
        }
    }
}

fun getPlayListWithSongs(userId: UUID, playlistId: UUID): Playlist? {
    return transaction {
        val playlistRow = Playlists.select { Playlists.id eq playlistId }.singleOrNull()
            ?: return@transaction null

        val songs = PlaylistSongs
            .leftJoin(Songs)
            .select { PlaylistSongs.playlistId eq playlistId }
            .map { row ->
                getSongWithArtistsAndFavorite(userId, row[Songs.id])!!
            }

        Playlist(
            id = playlistRow[Playlists.id],
            name = playlistRow[Playlists.name],
            coverUrl = null, // TODO: implement
            songsCount = songs.size,
            songs = Songs(
                count = songs.size,
                songs = songs
            )
        )
    }
}

fun getPlaylists(userId: UUID): ShortPlaylists {
    val playlists = transaction {
        Playlists
            .leftJoin(PlaylistSongs)
            .leftJoin(Songs)
            .slice(Playlists.fields + Songs.id.count())
            .select { Playlists.userId eq userId }
            .groupBy(*Playlists.fields.toTypedArray())
            .map {
                ShortPlaylist(
                    id = it[Playlists.id],
                    name = it[Playlists.name],
                    coverUrl = null, // TODO: implement
                    songsCount = it[Songs.id.count()].toInt(),
                )
            }
    }

    return ShortPlaylists(
        count = playlists.size,
        playlists = playlists
    )
}

fun createPlaylist(uuid: UUID, userId: UUID, request: CreatePlaylistRequest) {
    transaction {
        Playlists.insert {
            it[id] = uuid
            it[name] = request.name
            it[Playlists.userId] = userId
            it[createdAt] = DateTime.now()
        }
    }
}

fun deletePlaylist(userId: UUID, playlistId: UUID) {
    transaction {
        Playlists.deleteWhere { (Playlists.userId eq userId) and (Playlists.id eq playlistId) }
    }
}

fun updatePlaylist(playlistId: UUID, request: UpdatePlaylistRequest) {
    transaction {
        Playlists.update({ Playlists.id eq playlistId }) {
            it[name] = request.name
        }
    }
}