package com.witelokk.music.models

import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.Serializable
import java.util.UUID

@Serializable
data class CreatePlaylistResponse(
    @Serializable(UUIDSerializer::class) val id: UUID,
)