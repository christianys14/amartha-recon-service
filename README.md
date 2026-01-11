<h1>Amartha Recon Service</h1>

# Tech Stack:
1. Golang 1.25 (not necessary)
2. MySQL

# My Assumption
1. Amartha connected with one Aggregator. But, the aggregator could be connected with multiple bank.
2. The process reporting each bank, submit the transaction to the aggregator. Then aggregator will aggregate and report to Amartha.
3. When Amartha receive the report from aggregator, amartha system able to recon the transaction from a different bank. It means, only **single format** with contains multiple bank transaction.
4. Assume each file csv being uploaded, only allowed **max 100k rows**. But the value of 100k is configurable. Why we needed set max rows? In order to prevent from out of memory, or worse make it our system down.
5. From the requirement, that I know trxID != transaction_id, but the dummy data i was created trxID == transaction_id. Why? Because from my POV, it's pretty weird if we check only using amount only, and im confident enought thats not even the real case.
6. Since its only for test purpose, i didn't put any Authentications or Authorizations on the API. But, the real case we should applied, no matter what.
7. The csv system (amartha) the data is coming from big data which generated using BigQuery. But, example data attached here is dummy.
8. The dummy date generated range is 2026-01-01 to 2026-01-09.

# Structure Table Transaction
## Requirement
1. trxID
2. amount
3. type
4. transactionTime

## Example Data
| trxID | amount | type | transactionTime |
|-------|--------|------|-----------------|
| 12345 | 100.00 | DEBIT | 2023-10-01 12:00:00 |
| 67890 | 50.00  | CREDIT| 2023-10-02 14:30:00 |

## My Assumption
1. ID
2. transaction_id previously trxID
3. terminal_rrn
4. amount
5. transaction_type
6. bank_code
7. transaction_time previously date
8. updated_at

## Example Data
| ID | transaction_id | terminal_rrn | amount | transaction_type | bank_code | transaction_time |
|:---|:---|:---|:---|:---|:----------|:---|
| 1 | 12345 | RRN0019283 | 100.00 | DEBIT | 002       | 2023-10-01 12:00:00 |
| 2 | 67890 | RRN0019284 | 50.00 | CREDIT | 008       | 2023-10-02 14:30:00 |

## Explanations
1. Why we needed terminal_rrn? Because transaction_id is not equivalent with `unique_identifier` with bank. So, we need to add terminal_rrn to make it looks like "connected" between amartha system and bank. As far as i know, in order to recon, its should not only using field amount. I think we should use other fields such as Terminal RRN, Amount, Transaction Time. So we can guarantee that the transaction is valid and "same"
2. Why we needed bank_code? Because on table transaction, we can't identify this transaction should be posting into which bank? Further more how we should aggregate based on bank, if we don't have such information?
3. Why i've script sql on /db/migrations? Because in this test case, the example data is not yet available. So, i had to improve such create table, store procedure in order to generate dummy data.

# How to run
1. Clone the repository
2. Run `go run main.go serveHttp`, it will start the server on port 5051.
3. Here's the example curl command to upload the file csv:
> curl --location 'localhost:5051/v1/internal/recon' \
--form 'system=@"/amartha_transactions.csv"' \
--form 'bank=@"/amartha_transactions2.csv"' \
--form 'start_date="2026-01-01"' \
--form 'end_date="2026-01-03"'
4. file `amartha_transactions.csv` is dummy data for system (amartha).
5. file `amartha_transactions2.csv` is dummy data for bank. file `amartha_transactions2.csv` is part of `amartha_transactions.csv` with some of the data changed.
6. the bank statement mistmatched data example would be :
> 104235555421,5133.26,2026-01-03 00:00:00,002 <br>
> 104235574821,8022.26,2026-01-03 00:00:00,002

# Solution Approach
1. Distinct the transaction from bank_code.
2. Aggregate the transaction from amartha, and the bank statement based on bank_code.
3. Define max chunk.
4. Compare row on Transaction and Bank from row 1-10 on A, 11-20 on B, so on, so forth.
5. Then collect the result.

# Example
Imagine we have 2 bank_code, and we have chunk 4. Each chunk will compare between transaction and bank statement.
> bank_code 1: will process 20 rows at a time. BUT will divide based on the chunk.<br>
> bank_code 1, chunk A: proceed row 1-5.<br>
> bank_code 1, chunk B: proceed row 6-10.<br>
> bank_code 1, chunk C: proceed row 11-15.<br>
> bank_code 1, chunk D: proceed row 16-20.<br>
> so on forth<br>
> bank_code 2: will process 40 rows at a time. BUT, will divide based on the chunk.<br>
> bank_code 2, chunk A : proceed row 1-10.<br>
> bank_code 2, chunk B : proceed row 11–20.<br>
> bank_code 2, chunk C : proceed row 21-30.<br>
> bank_code 2, chunk D : proceed row 31-40.<br>
> So in total we have 8 chunks/routines spawned (count distinct(bank_code) * chunk).