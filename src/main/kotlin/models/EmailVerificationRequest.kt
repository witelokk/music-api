package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class EmailVerificationRequest(
    val email: String,
    val code: String,
)
