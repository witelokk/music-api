package com.witelokk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class StopFollowingRequest(
    @SerialName("artist_id") val artistId: String,
)
