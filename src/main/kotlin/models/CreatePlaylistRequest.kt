package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class CreatePlaylistRequest(
    val name: String,
)
