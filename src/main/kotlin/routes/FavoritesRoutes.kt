package com.witelokk.routes

import com.witelokk.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.PG_UNIQUE_VIOLATION
import com.witelokk.models.AddFavoriteSongRequest
import com.witelokk.models.FailureResponse
import com.witelokk.tables.Artists
import com.witelokk.tables.Favorites
import com.witelokk.tables.SongArtists
import com.witelokk.tables.Songs
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import models.ShortArtist
import models.Song
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.deleteWhere
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import java.sql.SQLException
import java.util.*


fun Route.favoriteRoutes() {
    authenticate("auth-jwt") {
        route("/favorites") {
            get("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val favorites = getFavoriteSongs(userId)

                call.respond(
                    com.witelokk.models.Songs(count = favorites.count(), songs = favorites),
                )
            }

            post("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<AddFavoriteSongRequest>()

                try {
                    addFavoriteSong(userId, UUID.fromString(request.songId))
                } catch (e: SQLException) {
                    if (e.sqlState == PG_FOREIGN_KEY_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("song_not_fount", "Song not found")
                        )
                    } else if (e.sqlState == PG_UNIQUE_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("already_favorite", "Song is already favorite")
                        )
                    }
                } catch (e: Exception) {
                    println(e)
                    return@post call.respond(
                        HttpStatusCode.InternalServerError,
                        FailureResponse("internal_error", "Internal error")
                    )
                }

                call.respond(HttpStatusCode.Created)
            }

            delete("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<AddFavoriteSongRequest>()

                try {
                    removeFavoriteSong(userId, UUID.fromString(request.songId))
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

fun getFavoriteSongs(userId: UUID): List<Song> {
    return transaction {
        val favoriteSongs = (Favorites innerJoin Songs)
            .select { Favorites.userId eq userId }
            .map { it[Songs.id] }

        favoriteSongs.mapNotNull { songId ->
            val songRow = Songs.select { Songs.id eq songId }.singleOrNull() ?: return@mapNotNull null

            val artists = (SongArtists innerJoin Artists)
                .select { SongArtists.songId eq songId }
                .map { artistRow ->
                    ShortArtist(
                        id = artistRow[Artists.id].toString(),
                        name = artistRow[Artists.name],
                        avatarUrl = artistRow[Artists.avatarUrl]
                    )
                }

            Song(
                id = songRow[Songs.id].toString(),
                name = songRow[Songs.name],
                coverUrl = songRow[Songs.coverUrl],
                isFavorite = true,
                durationSeconds = songRow[Songs.duration],
                artists = artists,
                streamUrl = songRow[Songs.streamUrl]
            )
        }
    }
}

fun addFavoriteSong(userId: UUID, songId: UUID) {
    transaction {
        Favorites.insert {
            it[Favorites.songId] = songId
            it[Favorites.userId] = userId
        }
    }
}

fun removeFavoriteSong(userId: UUID, songId: UUID) {
    transaction {
        Favorites.deleteWhere { Favorites.songId eq songId }
    }
}