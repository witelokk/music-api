package com.witelokk

import com.witelokk.routes.authRoutes
import com.witelokk.routes.userRoutes
import com.witelokk.routes.verificationRoutes
import io.github.crackthecodeabhi.kreds.connection.Endpoint
import io.github.crackthecodeabhi.kreds.connection.newClient
import io.ktor.server.application.*
import io.ktor.server.routing.*

fun main(args: Array<String>) {
    io.ktor.server.netty.EngineMain.main(args)
}

fun Application.module() {
    val databaseUrl = environment.config.property("database.url").getString()
    val databaseUser = environment.config.property("database.user").getString()
    val databasePassword = environment.config.property("database.password").getString()
    val redisUrl = environment.config.property("redis.url").getString()
    val jwtSecret = environment.config.property("jwt.secret").getString()

    connectToDatabase(databaseUrl, databaseUser, databasePassword)

    val redis = newClient(Endpoint.from(redisUrl))

    configureAuth(jwtSecret)
    configureSerialization()

    routing {
        userRoutes(redis)
        verificationRoutes(redis)
        authRoutes(redis, jwtSecret)
    }
}
