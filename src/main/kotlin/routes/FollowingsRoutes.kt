package com.witelokk.routes

import com.witelokk.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.PG_UNIQUE_VIOLATION
import com.witelokk.models.FailureResponse
import com.witelokk.models.StartFollowingRequest
import com.witelokk.models.StopFollowingRequest
import com.witelokk.tables.*
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import models.ShortArtist
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.and
import org.jetbrains.exposed.sql.deleteWhere
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import java.sql.SQLException
import java.util.*

fun Route.followingsRoutes() {
    authenticate("auth-jwt") {
        route("/followings") {
            get("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val followings = getFollowings(userId)

                call.respond(
                    com.witelokk.models.ShortArtists(count = followings.count(), artists = followings),
                )
            }

            post("/") {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<StartFollowingRequest>()

                try {
                    startFollowing(userId, UUID.fromString(request.artistId))
                } catch (e: SQLException) {
                    if (e.sqlState == PG_FOREIGN_KEY_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("artist_not_fount", "Artist not found")
                        )
                    } else if (e.sqlState == PG_UNIQUE_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("already_following", "You are already following this artist")
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
                val request = call.receive<StopFollowingRequest>()

                try {
                    stopFollowing(userId, UUID.fromString(request.artistId))
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

fun getFollowings(userId: UUID): List<ShortArtist> {
    return transaction {
        Followers.innerJoin(Artists)
            .select { Followers.userId eq userId }
            .map { ShortArtist(
                id = it[Artists.id].toString(),
                name = it[Artists.name],
                avatarUrl = it[Artists.avatarUrl],
            ) }
    }
}

fun startFollowing(userId: UUID, artistId: UUID) {
    transaction {
        Followers.insert {
            it[Followers.userId] = userId
            it[Followers.artistId] = artistId
        }
    }
}

fun stopFollowing(userId: UUID, artistId: UUID) {
    transaction {
        Followers.deleteWhere { (Followers.userId eq userId) and (Followers.artistId eq artistId) }
    }
}
