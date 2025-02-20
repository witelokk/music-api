package com.witelokk.models

import kotlinx.serialization.Serializable

@Serializable
data class FailureResponse(
    val error: String,
    val message: String,
)
