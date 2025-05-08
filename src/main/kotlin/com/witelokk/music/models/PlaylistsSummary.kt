package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class PlaylistsSummary(
    val count: Int,
    val playlists: List<PlaylistSummary>,
)
