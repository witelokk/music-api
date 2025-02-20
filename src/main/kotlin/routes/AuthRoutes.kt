package com.witelokk.routes

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
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
import java.time.temporal.ChronoUnit
import kotlin.time.Duration

val TOKEN_TTL = Duration.parse("24h")

fun Route.authRoutes(redis: KredsClient) {
    val secret = "secret"

    post("/tokens") {
        val request = call.receive<TokensRequest>()

        redis.use { client ->
            val isCodeValid = client.exists("verification:${request.email}:${request.code}") == 1L

            if (!isCodeValid) {
                return@post call.respond(HttpStatusCode.BadRequest, FailureResponse("invalid_code", "Invalid code"))
            }

            client.del("verification:${request.email}:${request.code}")
        }

        val user = transaction {
            Users.select { Users.email eq request.email }.singleOrNull()
        }

        if (user == null) {
            return@post call.respond(HttpStatusCode.BadRequest, FailureResponse("invalid_user", "User not found"))
        }

        val jwt = JWT.create()
            .withClaim("sub", user[Users.id].toString())
            .withExpiresAt(DateTime.now().plusSeconds(TOKEN_TTL.inWholeSeconds.toInt()).toDate())
            .sign(Algorithm.HMAC256("secret"))

        call.respond(HttpStatusCode.Created, TokensResponse(jwt))
    }
}
