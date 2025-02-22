package com.witelokk.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import com.google.api.client.googleapis.auth.oauth2.GoogleIdTokenVerifier
import com.witelokk.models.FailureResponse
import com.witelokk.models.TokensRequest
import com.witelokk.models.TokensResponse
import com.witelokk.tables.Users
import io.github.crackthecodeabhi.kreds.connection.KredsClient
import io.ktor.http.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import org.jetbrains.exposed.sql.select
import org.jetbrains.exposed.sql.transactions.transaction
import org.joda.time.DateTime
import kotlin.time.Duration


val TOKEN_TTL = Duration.parse("24h")

fun Route.authRoutes(redis: KredsClient, jwtSecret: String, googleIdTokenVerifier: GoogleIdTokenVerifier) {
    suspend fun validateCode(email: String, code: String): String? {
        redis.use { client ->
            val isCodeValid = client.exists("verification:${email}:${code}") == 1L

            if (!isCodeValid) {
                return null
            }

            client.del("verification:${email}:${code}")
            return email
        }

    }

    fun validateGoogleToken(token: String): String? {
        return googleIdTokenVerifier.verify(token)?.payload?.email
    }

    post("/tokens") {
        val request = call.receive<TokensRequest>()

        val email = when (request.grantType) {
            "code" -> {
                val email = request.email ?: return@post call.respond(
                    HttpStatusCode.BadRequest,
                    FailureResponse("no_email", "No email provided")
                )
                val code = request.code ?: return@post call.respond(
                    HttpStatusCode.BadRequest,
                    FailureResponse("no_code", "No code provided")
                )

                validateCode(email, code) ?: return@post call.respond(
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
                    FailureResponse("invalid_google_id_token", "Invalid google id")
                )
            }

            else -> {
                return@post call.respond(
                    HttpStatusCode.BadRequest,
                    FailureResponse("invalid_grant_type", "Invalid grant_type")
                )
            }
        }

        val user = transaction {
            Users.select { Users.email eq email }.singleOrNull()
        }

        if (user == null) {
            return@post call.respond(HttpStatusCode.BadRequest, FailureResponse("invalid_user", "User not found"))
        }

        val jwt = JWT.create()
            .withClaim("sub", user[Users.id].toString())
            .withExpiresAt(DateTime.now().plusSeconds(TOKEN_TTL.inWholeSeconds.toInt()).toDate())
            .sign(Algorithm.HMAC256(jwtSecret))

        call.respond(HttpStatusCode.Created, TokensResponse(jwt))
    }
}
