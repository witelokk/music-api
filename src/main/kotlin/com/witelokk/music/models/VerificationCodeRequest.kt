package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class VerificationCodeRequest (
    val email: String,
)
