package com.witelokk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class ShortPlaylist(
    val id: String,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    @SerialName("songs_count") val songsCount: Int,
)
