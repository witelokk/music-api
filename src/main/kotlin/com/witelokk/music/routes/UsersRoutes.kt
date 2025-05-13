package com.witelokk.music.routes

import com.witelokk.music.PG_FOREIGN_KEY_VIOLATION
import com.witelokk.music.models.CreateUserRequest
import com.witelokk.music.models.FailureResponse
import com.witelokk.music.models.User
import com.witelokk.music.tables.Users
import io.github.crackthecodeabhi.kreds.connection.KredsClient
import io.github.smiley4.ktoropenapi.get
import io.github.smiley4.ktoropenapi.post
import io.github.smiley4.ktoropenapi.route
import io.ktor.http.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.insert
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import org.joda.time.DateTime
import java.sql.SQLException
import java.util.*

fun Route.userRoutes(redis: KredsClient) {
    route("/users", {
        tags = listOf("users")
    }) {
        post({
            description = "Create a user"
            request {
                body<CreateUserRequest>()
            }
            response {
                HttpStatusCode.Created to {
                    description = "Success"
                }
                HttpStatusCode.BadRequest to {
                    description = "Bad request"
                    body<FailureResponse>()
                }
                HttpStatusCode.Conflict to {
                    description = "User already exists"
                    body<FailureResponse>()
                }
            }
        }) {
            val request = call.receive<CreateUserRequest>()

            redis.use { client ->
                val isCodeValid = client.exists("verification:${request.email}:${request.code}") == 1L

                if (!isCodeValid) {
                    return@post call.respond(HttpStatusCode.BadRequest, FailureResponse("invalid_code", "Invalid code"))
                }
            }

            // create a record in the database
            try {
                transaction {
                    Users.insert {
                        it[id] = UUID.randomUUID()
                        it[email] = request.email
                        it[name] = request.name
                        it[createdAt] = DateTime.now()
                    }
                }
            } catch (e: SQLException) {
                if (e.sqlState == PG_FOREIGN_KEY_VIOLATION) {
                    call.respond(
                        HttpStatusCode.Conflict, FailureResponse("user_exists", "User with such email already exists")
                    )
                } else {
                    call.respond(
                        HttpStatusCode.InternalServerError, FailureResponse("internal_error", "Internal error")
                    )
                }
            }

            call.respond(HttpStatusCode.Created)
        }

        authenticate("auth-jwt") {
            get("/me", {
                description = "Get info about current user"
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val user = transaction {
                    Users.select(Users.id eq userId).singleOrNull()
                }?.let { User(name = it[Users.name], email = it[Users.email]) }

                if (user == null) {
                    call.respond(HttpStatusCode.Unauthorized)
                    return@get
                }

                call.respond(HttpStatusCode.OK, user)
            }
        }
    }
}