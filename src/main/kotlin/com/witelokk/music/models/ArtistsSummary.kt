package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class ArtistsSummary(
    val count: Int,
    val artists: List<ArtistSummary>,
) {
    val names = artists.joinToString(", ") { it.name }
}
