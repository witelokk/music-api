package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class Releases(
    val count: Int,
    val releases: List<Release>
)
