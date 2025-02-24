package com.witelokk.music

import com.google.api.client.googleapis.auth.oauth2.GoogleIdTokenVerifier
import com.google.api.client.http.javanet.NetHttpTransport
import com.google.api.client.json.gson.GsonFactory
import com.witelokk.music.routes.*
import io.github.crackthecodeabhi.kreds.connection.Endpoint
import io.github.crackthecodeabhi.kreds.connection.newClient
import io.github.smiley4.ktorswaggerui.routing.openApiSpec
import io.github.smiley4.ktorswaggerui.routing.swaggerUI
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
    val googleAuthAudience = environment.config.property("google-auth.audience").getString().split(',')

    val emailSender: EmailSender = MailgunEmailSender(mailgunApiKey, mailGunDomain, mailGunFrom)

    connectToDatabase(databaseUrl, databaseUser, databasePassword)

    val redis = newClient(Endpoint.from(redisUrl))

    configureAuth(jwtSecret)
    configureSerialization()
    configureSwagger()
    configureLogging()

    val tokenVerifier = GoogleIdTokenVerifier.Builder(NetHttpTransport(), GsonFactory())
        .setAudience(googleAuthAudience)
        .build()

    routing {
        userRoutes(redis)
        verificationRoutes(redis, emailSender)
        authRoutes(redis, jwtSecret, tokenVerifier)

        artistsRoutes()
        songsRoutes()
        favoriteRoutes()
        followingsRoutes()
        releasesRoutes()
        playlistsRoutes()
        playlistSongsRoutes()

        route("api.json") {
            openApiSpec()
        }
        route("swagger") {
            swaggerUI("/api.json")
        }
    }
}
