package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class SearchResultItem(
    val type: String,
    val song: Song? = null,
    val release: ShortRelease? = null,
    val artist: ShortArtist? = null,
    val playlist: ShortPlaylist? = null,
)
