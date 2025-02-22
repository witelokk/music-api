package com.witelokk.models

import kotlinx.serialization.Serializable
import models.ShortArtist

@Serializable
data class ShortArtists(
    val count: Int,
    val artists: List<ShortArtist>
)
