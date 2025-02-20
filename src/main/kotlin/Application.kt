package com.witelokk

import io.github.crackthecodeabhi.kreds.connection.Endpoint
import io.github.crackthecodeabhi.kreds.connection.newClient
import io.ktor.server.application.*

fun main(args: Array<String>) {
    io.ktor.server.netty.EngineMain.main(args)
}

fun Application.module() {
    connectToDatabase()

    val redis = newClient(Endpoint.from("127.0.0.1:6377"))

    configureAuth()
    configureSerialization()
    configureRouting(redis)
}
