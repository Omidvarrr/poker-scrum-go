-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS join_requests;
DROP TABLE IF EXISTS room_members;
-- +goose StatementEnd