package com.witelokk.music.models

import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.*

@Serializable
data class AddSongToPlaylistRequest(
    @SerialName("song_id") @Serializable(UUIDSerializer::class) val songId: UUID,
)
