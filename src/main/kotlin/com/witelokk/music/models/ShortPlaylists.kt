package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class ShortPlaylists(
    val count: Int,
    val playlists: List<ShortPlaylist>,
)
