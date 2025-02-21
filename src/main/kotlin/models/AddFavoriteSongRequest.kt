package com.witelokk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class AddFavoriteSongRequest(
    @SerialName("song_id") val songId: String,
)
