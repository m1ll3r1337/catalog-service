package testutil

import "github.com/gofrs/uuid"

func PtrString(s string) *string { return &s }

func PtrInt64(v int64) *int64 { return &v }

func PtrUUID(v uuid.UUID) *uuid.UUID { return &v }
