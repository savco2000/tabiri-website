package main

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.ErrorContext(
		r.Context(),
		"request failed",
		"error", err,
		"request_id", requestIDFromContext(r.Context()),
		"method", r.Method,
		"route", r.Pattern,
	)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	name string,
	data PageData,
) {
	ctx, span := app.telemetry.tracer.Start(
		r.Context(),
		"render "+name,
		trace.WithAttributes(attribute.String("template.name", name)),
	)
	defer span.End()

	tmpl, ok := app.templates.pages[name]
	if !ok {
		app.telemetry.templateFailures.Add(ctx, 1, metric.WithAttributes(
			attribute.String("template.name", name),
		))
		span.SetStatus(codes.Error, "template not cached")
		app.serverError(w, r, fmt.Errorf("page template %q is not cached", name))
		return
	}
	output, err := executeTemplate(tmpl, data)
	if err != nil {
		app.telemetry.templateFailures.Add(ctx, 1, metric.WithAttributes(
			attribute.String("template.name", name),
		))
		span.RecordError(err)
		span.SetStatus(codes.Error, "template rendering failed")
		app.serverError(w, r, fmt.Errorf("render page template %q: %w", name, err))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(output); err != nil {
		app.logger.ErrorContext(r.Context(), "write response", "error", err)
	}
}

func (app *application) databaseOperation(
	ctx context.Context,
	operation string,
	fn func(context.Context) error,
) error {
	ctx, span := app.telemetry.tracer.Start(
		ctx,
		"database "+operation,
		trace.WithAttributes(attribute.String("db.operation.name", operation)),
	)
	defer span.End()
	started := app.now()
	err := fn(ctx)
	app.telemetry.recordDatabase(ctx, operation, app.now().Sub(started), err)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "database operation failed")
	}
	return err
}
