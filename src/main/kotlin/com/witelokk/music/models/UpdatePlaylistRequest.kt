package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class UpdatePlaylistRequest(
    val name: String,
)
