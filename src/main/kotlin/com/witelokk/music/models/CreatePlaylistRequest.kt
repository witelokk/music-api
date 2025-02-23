package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class CreatePlaylistRequest(
    val name: String,
)
