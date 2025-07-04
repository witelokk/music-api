package com.witelokk.music.routes

import com.witelokk.music.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.music.PG_UNIQUE_VIOLATION
import com.witelokk.music.models.FailureResponse
import com.witelokk.music.models.StartFollowingRequest
import com.witelokk.music.models.StopFollowingRequest
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
import com.witelokk.music.models.ArtistsSummary
import com.witelokk.music.tables.Artists
import com.witelokk.music.tables.Followers
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.transactions.transaction
import java.sql.SQLException
import java.util.*

fun Route.followingsRoutes() {
    authenticate("auth-jwt") {
        route("/followings", {
            tags = listOf("followings")
        }) {
            get({
                description = "Get a list of followed artists"
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<ArtistsSummary>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val followings = getFollowings(userId)

                call.respond(
                    ArtistsSummary(count = followings.count(), artists = followings),
                )
            }

            post({
                description = "Start following an artist"
                request {
                    body<StartFollowingRequest>()
                }
                response {
                    HttpStatusCode.Created to {
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
                val request = call.receive<StartFollowingRequest>()

                try {
                    startFollowing(userId, request.artistId)
                } catch (e: SQLException) {
                    if (e.sqlState == PG_FOREIGN_KEY_VIOLATION) {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("artist_not_found", "Artist not found")
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

            delete({
                description = "Stop following an artist"
                request {
                    body<StopFollowingRequest>()
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())
                val request = call.receive<StopFollowingRequest>()

                try {
                    stopFollowing(userId, request.artistId)
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

fun getFollowings(userId: UUID): List<ArtistSummary> {
    return transaction {
        (Followers innerJoin Artists)
            .slice(
                Artists.id,
                Artists.name,
                Artists.avatarUrl,
                Artists.coverUrl,
                Followers.artistId.count() // Count followers per artist
            )
            .select { Followers.artistId inSubQuery
                    Followers.slice(Followers.artistId)
                        .select { Followers.userId eq userId }
            }
            .groupBy(Artists.id, Artists.name, Artists.avatarUrl, Artists.coverUrl)
            .map {
                ArtistSummary(
                    id = it[Artists.id],
                    name = it[Artists.name],
                    avatarUrl = it[Artists.avatarUrl],
                )
            }
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
