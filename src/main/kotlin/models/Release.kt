package com.witelokk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Release(
    val id: String,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    val type: String,
    @SerialName("released_at") val releasedAt: String,
    val songs: Songs,
    val artists: ShortArtists,
)
