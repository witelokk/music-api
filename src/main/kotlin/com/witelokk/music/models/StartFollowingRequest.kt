package com.witelokk.music.models

import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.UUID

@Serializable
data class StartFollowingRequest(
    @SerialName("artist_id") @Serializable(UUIDSerializer::class) val artistId: UUID,
)
