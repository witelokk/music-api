package com.witelokk.music

interface EmailSender {
    suspend fun sendEmail(to: List<String>, subject: String, text: String)
}
