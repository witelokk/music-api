package com.witelokk

import io.ktor.client.*
import io.ktor.client.engine.cio.*
import io.ktor.client.request.*
import io.ktor.client.request.forms.*
import io.ktor.client.statement.*
import io.ktor.http.*

class MailgunEmailSender(
    private val apiKey: String,
    private val domain: String,
    private val from: String,
): EmailSender {
    private val client = HttpClient(CIO)

    override suspend fun sendEmail(to: List<String>, subject: String, text: String) {
        val url = "https://api.mailgun.net/v3/$domain/messages"

        val response: HttpResponse = client.submitForm(
            url = url,
            formParameters = Parameters.build {
                append("from", from)
                appendAll("to", to)
                append("subject", subject)
                append("text", text)
            }
        ) {
            headers {
                append(HttpHeaders.Authorization, "Basic ${encodeCredentials(apiKey)}")
            }
        }

        if (!response.status.isSuccess()) {
            throw Exception(response.readRawBytes().decodeToString())
        }
    }

    private fun encodeCredentials(apiKey: String): String {
        // Encode the API key in the format "api:YOUR_API_KEY" using Base64
        val credentials = "api:$apiKey"
        return java.util.Base64.getEncoder().encodeToString(credentials.toByteArray())
    }
}

