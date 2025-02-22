package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class UpdatePlaylistRequest(
    val name: String,
)
