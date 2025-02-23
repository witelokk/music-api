package com.witelokk.routes

import com.witelokk.EmailSender
import com.witelokk.MailgunEmailSender
import com.witelokk.models.FailureResponse
import com.witelokk.models.VerificationCodeRequest
import io.github.crackthecodeabhi.kreds.connection.Endpoint
import io.github.crackthecodeabhi.kreds.connection.KredsClient
import io.github.crackthecodeabhi.kreds.connection.newClient
import io.github.smiley4.ktorswaggerui.dsl.routing.post
import io.ktor.http.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import org.joda.time.DateTime
import kotlin.math.max
import kotlin.time.Duration

val SEND_NEW_CODE_AFTER = Duration.parse("2m")
val CODE_TTL = Duration.parse("10m")

fun Route.verificationRoutes(redis: KredsClient, emailSender: EmailSender) {
    post("/verification-code-request", {
        tags = listOf("auth")
        description = "Request a verification code"
        request {
            body<VerificationCodeRequest>()
        }
        response {
            HttpStatusCode.Created to {
                description = "Success"
            }
            HttpStatusCode.TooManyRequests to {
                description = "Too many verification requests"
                body<FailureResponse>()
            }
        }
    }) {
        val request = call.receive<VerificationCodeRequest>()

        // generate and save verification code
        val code = (1000..9999).random().toString()

        redis.use { client ->
            val keys = client.keys("verification:${request.email}:*")
            var mostRecentCodeTime = 0L
            for (key in keys) {
                mostRecentCodeTime = max(mostRecentCodeTime, client.get(key)!!.toLong())
            }

            if (keys.isNotEmpty() && mostRecentCodeTime < DateTime.now().millis + SEND_NEW_CODE_AFTER.inWholeMilliseconds) {
                return@post call.respond(
                    HttpStatusCode.TooManyRequests,
                    FailureResponse("too_many_verification_requests", "Too many verification requests. Try again later")
                )
            }

            client.set("verification:${request.email}:${code}", DateTime.now().millis.toString())
            client.expire("verification:${request.email}:${code}", CODE_TTL.inWholeSeconds.toULong())
        }

        // send verification email
        try {
            emailSender.sendEmail(
                listOf(request.email), "Account Verification", """Your Music code: <strong>${code}</strong>
                    |
                    |The code will expire in ${CODE_TTL.inWholeMinutes} minutes""".trimMargin()
            )
        } catch (e: Exception) {
            println(e.message)
            return@post call.respond(
                HttpStatusCode.InternalServerError, FailureResponse("internal_error", "Internal error")
            )
        }

        call.respond(HttpStatusCode.Created)
    }
}