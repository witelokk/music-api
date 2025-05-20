package com.witelokk.music.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class TokensRequest(
    @SerialName("grant_type") val grantType: String,
    val email: String? = null,
    val code: String? = null,
    @SerialName("google_token") val googleToken: String? = null,
    @SerialName("refresh_token") val refreshToken: String? = null,
)
