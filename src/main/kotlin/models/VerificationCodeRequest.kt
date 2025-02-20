package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class VerificationCodeRequest (
    val email: String,
)
