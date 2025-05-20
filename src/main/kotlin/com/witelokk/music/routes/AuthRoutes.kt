package com.witelokk.music.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import com.google.api.client.googleapis.auth.oauth2.GoogleIdTokenVerifier
import com.witelokk.music.models.FailureResponse
import com.witelokk.music.models.TokensRequest
import com.witelokk.music.models.TokensResponse
import com.witelokk.music.tables.Users
import io.github.crackthecodeabhi.kreds.connection.KredsClient
import io.github.smiley4.ktoropenapi.post
import io.ktor.http.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.transactions.transaction
import org.joda.time.DateTime
import java.util.*
import kotlin.time.Duration


val TOKEN_TTL = Duration.parse("24h")
val REFRESH_TOKEN_TTL = Duration.parse("30d")

fun generateRefreshToken(): String = UUID.randomUUID().toString()

fun Route.authRoutes(redis: KredsClient, jwtSecret: String, googleIdTokenVerifier: GoogleIdTokenVerifier) {
    suspend fun validateCode(email: String, code: String): String? {
        redis.use { client ->
            val isCodeValid = client.exists("verification:${email}:${code}") == 1L
            if (!isCodeValid) {
                return null
            }
            return email
        }
    }

    fun validateGoogleToken(token: String): Pair<String, String?>? {
        val payload = googleIdTokenVerifier.verify(token)?.payload ?: return null
        return payload.email to payload["name"] as? String
    }

    post("/tokens", {
        tags = listOf("auth")
        description = "Get tokens"
        request {
            body<TokensRequest>()
        }
        response {
            HttpStatusCode.Created to {
                description = "Success"
                body<TokensResponse>()
            }
            HttpStatusCode.BadRequest to {
                description = "Bad request"
                body<FailureResponse>()
            }
        }
    }) {
        val request = call.receive<TokensRequest>()

        when (request.grantType) {
            "code", "google_token" -> {
                val (email, name) = when (request.grantType) {
                    "code" -> {
                        val email = request.email ?: return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("no_email", "No email provided")
                        )
                        val code = request.code ?: return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("no_code", "No code provided")
                        )
                        validateCode(email, code)?.let { it to null } ?: return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("invalid_code", "Invalid code")
                        )
                    }

                    "google_token" -> {
                        val token = request.googleToken ?: return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("no_google_token", "No google_token provided")
                        )
                        validateGoogleToken(token) ?: return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("invalid_google_id_token", "Invalid Google ID token")
                        )
                    }

                    else -> error("unreachable")
                }

                val user = transaction {
                    Users.select { Users.email eq email }.singleOrNull()
                }

                if (user == null) {
                    if (request.grantType == "google_token") {
                        transaction {
                            val newUserId = UUID.randomUUID()
                            val newName = name ?: "Unknown User"
                            Users.insert {
                                it[id] = newUserId
                                it[Users.name] = newName
                                it[Users.email] = email
                                it[createdAt] = DateTime.now()
                            }
                        }
                    } else {
                        return@post call.respond(
                            HttpStatusCode.BadRequest,
                            FailureResponse("invalid_user", "User not found")
                        )
                    }
                }

                val existingUser = transaction {
                    Users.select { Users.email eq email }.single()
                }

                val jwt = JWT.create()
                    .withClaim("sub", existingUser[Users.id].toString())
                    .withExpiresAt(DateTime.now().plusSeconds(TOKEN_TTL.inWholeSeconds.toInt()).toDate())
                    .sign(Algorithm.HMAC256(jwtSecret))

                val refreshToken = generateRefreshToken()
                redis.use { client ->
                    client.set("refresh_token:$refreshToken", existingUser[Users.id].toString())
                    client.expire("refresh_token:$refreshToken", REFRESH_TOKEN_TTL.inWholeSeconds.toULong())
                }

                call.respond(HttpStatusCode.Created, TokensResponse(jwt, refreshToken))
            }

            "refresh_token" -> {
                val refreshToken = request.refreshToken ?: return@post call.respond(
                    HttpStatusCode.BadRequest,
                    FailureResponse("no_refresh_token", "No refresh token provided")
                )
                val userId = redis.use { client ->
                    val value = client.get("refresh_token:$refreshToken")
                    if (value == null) {
                        null
                    } else {
                        client.del("refresh_token:$refreshToken")
                        value
                    }
                } ?: return@post call.respond(
                    HttpStatusCode.BadRequest,
                    FailureResponse("invalid_refresh_token", "Invalid or expired refresh token")
                )

                val jwt = JWT.create()
                    .withClaim("sub", userId)
                    .withExpiresAt(DateTime.now().plusSeconds(TOKEN_TTL.inWholeSeconds.toInt()).toDate())
                    .sign(Algorithm.HMAC256(jwtSecret))

                val newRefreshToken = generateRefreshToken()
                redis.use { client ->
                    client.set("refresh_token:$newRefreshToken", userId)
                    client.expire("refresh_token:$newRefreshToken", REFRESH_TOKEN_TTL.inWholeSeconds.toULong())
                }

                call.respond(HttpStatusCode.Created, TokensResponse(jwt, newRefreshToken))
            }

            else -> return@post call.respond(
                HttpStatusCode.BadRequest,
                FailureResponse("invalid_grant_type", "Invalid grant_type")
            )
        }
    }
}