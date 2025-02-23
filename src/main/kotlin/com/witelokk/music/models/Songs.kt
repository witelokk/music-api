package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class Songs(
    val count: Int,
    val songs: List<Song>
)
