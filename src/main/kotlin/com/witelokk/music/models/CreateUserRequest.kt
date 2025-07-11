package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class CreateUserRequest(
    val email: String,
    val name: String,
    val code: String,
)
