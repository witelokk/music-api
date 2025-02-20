package com.witelokk

interface EmailSender {
    suspend fun sendEmail(to: List<String>, subject: String, text: String)
}
