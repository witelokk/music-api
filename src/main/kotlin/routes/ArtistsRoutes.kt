package com.witelokk.routes

import com.witelokk.models.FailureResponse
import com.witelokk.tables.Artists
import com.witelokk.tables.Followers
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import models.Artist
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.*

fun Route.artistsRoutes() {
    authenticate("auth-jwt") {
        route("/artists") {
            get("/{id}") {
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
                    getAritstWithFollowing(artistId, userId)
                }

                if (artist == null) {
                    println("artist == null")
                    return@get call.respond(
                        HttpStatusCode.NotFound, FailureResponse("artist_not_found", "Artist not found123")
                    )
                }

                return@get call.respond(artist)
            }
        }
    }
}

private fun getAritstWithFollowing(artistId: UUID, userId: UUID): Artist? {
    val followingAlias = exists(
        Followers.select {
            (Followers.artistId eq artistId) and (Followers.userId eq userId)
        }
    ).alias("following")

    return Artists.leftJoin(Followers).slice(
        Artists.id, Artists.name, Artists.coverUrl, Artists.avatarUrl, Followers.userId.count(), followingAlias
    ).select { Artists.id eq artistId }
        .groupBy(Artists.id, Artists.name, Artists.coverUrl, Artists.avatarUrl).map {
            Artist(
                id = it[Artists.id].toString(),
                name = it[Artists.name],
                avatarUrl = it[Artists.avatarUrl],
                coverUrl = it[Artists.coverUrl],
                followers = it[Followers.userId.count()].toInt() ?: 0,
                following = it[followingAlias]
            )
        }.singleOrNull()
}