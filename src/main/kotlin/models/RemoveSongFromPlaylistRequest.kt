package com.witelokk.models

import com.witelokk.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.UUID

@Serializable
data class RemoveSongFromPlaylistRequest(
    @SerialName("song_id") @Serializable(UUIDSerializer::class) val songId: UUID,
)
