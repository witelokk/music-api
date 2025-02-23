package com.witelokk.models

import com.witelokk.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.*

@Serializable
data class RemoveFavoriteSongRequest(
    @SerialName("song_id") @Serializable(UUIDSerializer::class) val songId: UUID,
)
