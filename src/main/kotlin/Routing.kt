package com.witelokk

import com.witelokk.routes.authRoutes
import com.witelokk.routes.userRoutes
import com.witelokk.routes.verificationRoutes
import io.github.crackthecodeabhi.kreds.connection.KredsClient
import io.ktor.server.application.*
import io.ktor.server.routing.*

fun Application.configureRouting(redis: KredsClient) {
    routing {
        userRoutes(redis)
        verificationRoutes(redis)
        authRoutes(redis)
    }
}
