package com.witelokk.music.routes

import com.witelokk.music.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.music.PG_UNIQUE_VIOLATION
import com.witelokk.music.models.*
import com.witelokk.music.tables.Artists
import com.witelokk.music.tables.Favorites
import com.witelokk.music.tables.SongArtists
import com.witelokk.music.tables.Songs
import io.github.smiley4.ktoropenapi.delete
import io.github.smiley4.ktoropenapi.get
import io.github.smiley4.ktoropenapi.post
import io.github.smiley4.ktoropenapi.route
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import com.witelokk.music.models.ArtistSummary
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.and
import org.jetbrains.exposed.sql.deleteWhere
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import java.sql.SQLException
import java.util.*


fun Route.favoriteRoutes() {
    authenticate("auth-jwt") {
        route("/favorites", {
            tags = listOf("favorites")
        }) {
            get({
                description = "Get favorite songs"
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<com.witelokk.music.models.Songs>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val favorites = getFavoriteSongs(userId)

                call.respond(
                    Songs(count = favorites.count(), songs = favorites),
                )
            }

            post({
                description = "Add favorite song"
                request {
                    body<AddFavoriteSongRequest>()
                }
                response {
                    HttpStatusCode.Created to {
                        description = "Success"
                    }
                    HttpStatusCode.BadRequest to {
                        description = "Bad Request"
                        body<FailureResponse>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<AddFavoriteSongRequest>()

                try {
                    addFavoriteSong(userId, request.songId)
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

            delete({
                description = "Remove favorite song"
                request {
                    body<RemoveFavoriteSongRequest>()
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<RemoveFavoriteSongRequest>()

                try {
                    removeFavoriteSong(userId, request.songId)
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
                    ArtistSummary(
                        id = artistRow[Artists.id],
                        name = artistRow[Artists.name],
                        avatarUrl = artistRow[Artists.avatarUrl]
                    )
                }

            Song(
                id = songRow[Songs.id],
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
        Favorites.deleteWhere { (Favorites.songId eq songId) and (Favorites.userId eq userId) }
    }
}