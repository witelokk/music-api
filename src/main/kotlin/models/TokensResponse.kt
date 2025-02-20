package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class TokensResponse(
    val accessToken: String,
)