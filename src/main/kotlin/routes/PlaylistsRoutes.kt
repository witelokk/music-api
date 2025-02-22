package com.witelokk.routes

import com.witelokk.models.*
import com.witelokk.tables.Playlists
import com.witelokk.tables.PlaylistsSongs
import com.witelokk.tables.Songs
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
        route("/playlists") {
            get("/{id}") {
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

            get("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val playlists = getPlaylistsWithSongs(userId)

                call.respond(playlists)
            }

            post("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<CreatePlaylistRequest>()

                val id = UUID.randomUUID()
                createPlaylist(id, userId, request)
                call.respond(HttpStatusCode.Created, CreatePlaylistResponse(id.toString()))
            }

            delete("/{id}") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val playlistId = try {
                    UUID.fromString(call.parameters["id"])
                } catch (e: IllegalArgumentException) {
                    return@delete call.respond(
                        HttpStatusCode.NotFound, FailureResponse("playlist_not_found", "Playlist not found")
                    )
                }

                deletePlaylist(playlistId, userId)
            }

            put("/id") {
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

                if (playlistUserId == null || playlistUserId != userId) {
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

        val songs = PlaylistsSongs
            .leftJoin(Songs)
            .select { PlaylistsSongs.playlistId eq playlistId }
            .map { row ->
                getSongWithArtistsAndFavorite(userId, row[Songs.id])!!
            }

        Playlist(
            id = playlistRow[Playlists.id].toString(),
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

fun getPlaylistsWithSongs(userId: UUID): com.witelokk.models.ShortPlaylists {
    val playlists = transaction {
        Playlists
            .leftJoin(PlaylistsSongs)
            .leftJoin(Songs)
            .slice(Playlists.fields + Songs.id.count())
            .select { Playlists.userId eq userId }
            .groupBy(*Playlists.fields.toTypedArray())
            .map {
                ShortPlaylist(
                    id = it[Playlists.id].toString(),
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