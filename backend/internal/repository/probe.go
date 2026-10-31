package repository

import "context"

type Probe interface{ Check(context.Context) error }
