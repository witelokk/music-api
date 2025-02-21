package com.witelokk.models

import kotlinx.serialization.Serializable
import models.Song

@Serializable
data class Songs(
    val count: Int,
    val songs: List<Song>
)
