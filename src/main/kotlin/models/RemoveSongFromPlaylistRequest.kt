package com.witelokk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class RemoveSongFromPlaylistRequest(
    @SerialName("song_id") val songId: String,
)
