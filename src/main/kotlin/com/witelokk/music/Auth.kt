package com.witelokk.music

import com.auth0.jwt.*
import com.auth0.jwt.algorithms.*
import com.witelokk.music.models.FailureResponse
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.response.*


fun Application.configureAuth(secret: String) {
    install(Authentication) {
        jwt("auth-jwt") {
            verifier(
                JWT.require(Algorithm.HMAC256(secret)).build()
            )
            validate { credential ->
                if (credential.payload.getClaim("username").asString() != "") {
                    JWTPrincipal(credential.payload)
                } else {
                    null
                }
            }
            challenge { defaultScheme, realm ->
                call.respond(
                    HttpStatusCode.Unauthorized,
                    FailureResponse("unauthorized", "Token is not valid or has expired")
                )
            }

        }
    }
}