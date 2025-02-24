package com.witelokk.music

import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.*
import io.ktor.server.plugins.calllogging.*
import io.ktor.server.request.*
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import org.slf4j.event.Level

@Serializable
data class LogData(
    val timestamp: Long,
    val ip: String,
    val method: String,
    val path: String,
    val userAgent: String,
    val requestId: String,
    val statusCode: Int?,
)

fun Application.configureLogging() {
    install(CallLogging) {
        level = Level.INFO
        format { call ->
            val statusCode = call.response.status()?.value

            val logData = LogData(
                timestamp = System.currentTimeMillis(),
                ip = call.request.origin.remoteHost,
                method = call.request.httpMethod.value,
                path = call.request.uri,
                userAgent = call.request.headers["User-Agent"] ?: "Unknown",
                requestId = call.request.headers["X-Request-ID"] ?: "No Request ID",
                statusCode = statusCode,
            )
            Json.encodeToString(logData)
        }
    }
}