package com.witelokk.music

import com.google.api.client.googleapis.auth.oauth2.GoogleIdTokenVerifier
import com.google.api.client.http.javanet.NetHttpTransport
import com.google.api.client.json.gson.GsonFactory
import com.witelokk.music.routes.*
import io.github.crackthecodeabhi.kreds.connection.Endpoint
import io.github.crackthecodeabhi.kreds.connection.newClient
import io.github.smiley4.ktoropenapi.OpenApiPlugin
import io.github.smiley4.ktoropenapi.config.OpenApiPluginConfig
import io.github.smiley4.ktoropenapi.config.OutputFormat
import io.github.smiley4.ktoropenapi.route
import io.github.smiley4.ktorswaggerui.swaggerUI
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.response.*
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
    val mailGunRegion = when (environment.config.property("mailgun.region").getString()) {
        "eu" -> MailGunRegion.EU
        "us" -> MailGunRegion.US
        else -> {
            throw RuntimeException("MailGun region must be either 'eu' or 'us'")
        }
    }
    val googleAuthAudience = environment.config.property("google-auth.audience").getString().split(',')

    val emailSender: EmailSender = MailgunEmailSender(mailgunApiKey, mailGunDomain, mailGunFrom, mailGunRegion)

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
        searchRoutes()
        homeScreenLayoutRoutes()

        route("api.json", { hidden = true }) {
            get {
                val contentType = when (OpenApiPlugin.getOpenApiSpecFormat(OpenApiPluginConfig.DEFAULT_SPEC_ID)) {
                    OutputFormat.JSON -> ContentType.Application.Json
                    OutputFormat.YAML -> ContentType.Text.Plain
                }

                val openApiSpec =
                    OpenApiPlugin.getOpenApiSpec(OpenApiPluginConfig.DEFAULT_SPEC_ID)
                        .replace("com.witelokk.music.models.", "")

                call.respondText(contentType, HttpStatusCode.OK) { openApiSpec }
            }
        }

        route("swagger") {
            swaggerUI("/api.json")
        }
    }
}
