package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class TokensRequest(
    val email: String,
    val code: String,
)
