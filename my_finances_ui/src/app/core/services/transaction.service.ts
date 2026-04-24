import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Transaction, PaginatedResponse } from '../models/models';

@Injectable({ providedIn: 'root' })
export class TransactionService {
  private http = inject(HttpClient);

  list(bankAccountId?: string, page = 1, limit = 20): Observable<PaginatedResponse<Transaction>> {
    let params = new HttpParams().set('page', page).set('limit', limit);
    if (bankAccountId) params = params.set('bank_account_id', bankAccountId);
    return this.http.get<PaginatedResponse<Transaction>>('/api/transactions', { params });
  }

  create(data: {
    name: string; value: number; operation: 'add' | 'subtract';
    bank_account_id: string; date: string; tag_ids: string[];
  }): Observable<Transaction> {
    return this.http.post<Transaction>('/api/transactions', data);
  }

  update(id: string, data: {
    name: string; value: number; operation: 'add' | 'subtract'; date: string; tag_ids: string[];
  }): Observable<Transaction> {
    return this.http.put<Transaction>(`/api/transactions/${id}`, data);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`/api/transactions/${id}`);
  }
}
