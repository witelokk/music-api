package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class SearchResultItem(
    val type: String,
    val song: Song? = null,
    val release: ReleaseSummary? = null,
    val artist: ArtistSummary? = null,
    val playlist: PlaylistSummary? = null,
)
