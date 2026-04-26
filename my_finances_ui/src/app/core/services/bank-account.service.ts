import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { BankAccount, StatementResponse } from '../models/models';

@Injectable({ providedIn: 'root' })
export class BankAccountService {
  private http = inject(HttpClient);

  list(): Observable<BankAccount[]> {
    return this.http.get<BankAccount[]>('/api/bank-accounts');
  }

  get(id: string): Observable<BankAccount> {
    return this.http.get<BankAccount>(`/api/bank-accounts/${id}`);
  }

  create(name: string, initialBalance: number, iconUrl?: string): Observable<BankAccount> {
    return this.http.post<BankAccount>('/api/bank-accounts', {
      name,
      initial_balance: initialBalance,
      icon_url: iconUrl ?? null,
    });
  }

  update(id: string, name: string, iconUrl?: string): Observable<BankAccount> {
    return this.http.put<BankAccount>(`/api/bank-accounts/${id}`, {
      name,
      icon_url: iconUrl ?? null,
    });
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`/api/bank-accounts/${id}`);
  }

  statement(id: string, page = 1, limit = 20): Observable<StatementResponse> {
    const params = new HttpParams().set('page', page).set('limit', limit);
    return this.http.get<StatementResponse>(`/api/bank-accounts/${id}/statement`, { params });
  }
}
