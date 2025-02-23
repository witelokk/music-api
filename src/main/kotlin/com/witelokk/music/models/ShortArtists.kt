package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class ShortArtists(
    val count: Int,
    val artists: List<ShortArtist>,
) {
    val names = artists.joinToString(", ") { it.name }
}
