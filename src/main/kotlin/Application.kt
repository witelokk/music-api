package com.witelokk

import com.witelokk.routes.*
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
    val mailgunApiKey = environment.config.property("mailgun.api_key").getString()
    val mailGunDomain = environment.config.property("mailgun.domain").getString()
    val mailGunFrom = environment.config.property("mailgun.from").getString()

    val emailSender: EmailSender = MailgunEmailSender(mailgunApiKey, mailGunDomain, mailGunFrom)

    connectToDatabase(databaseUrl, databaseUser, databasePassword)

    val redis = newClient(Endpoint.from(redisUrl))

    configureAuth(jwtSecret)
    configureSerialization()

    routing {
        userRoutes(redis)
        verificationRoutes(redis, emailSender)
        authRoutes(redis, jwtSecret)

        artistsRoutes()
        songsRoutes()
        favoriteRoutes()
    }
}
