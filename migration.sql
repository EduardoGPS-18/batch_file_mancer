CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE bank_slip_file (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);


CREATE TABLE bank_slip (
  debt_id PRIMARY KEY UNIQUE UUID,
  debt_amount NUMERIC(10,2) NOT NULL,
  debt_due_date DATE NOT NULL,
  user_name VARCHAR(255) NOT NULL,
  government_id INT NOT NULL,
  user_email VARCHAR(255) NOT NULL,
  bank_slip_file_id UUID NOT NULL,
  error_message varchar(255),
  status VARCHAR(50) NOT NULL,
  FOREIGN KEY (bank_slip_file_id) REFERENCES bank_slip_file(id),
  CONSTRAINT status_check CHECK (status IN ('PENDING', 'SUCCESS', 'GENERATING_BILLING_ERROR', 'SENT_EMAIL_WITH_ERROR'))
);

CREATE INDEX bank_slip_debt_id_idx ON bank_slip(debt_id);