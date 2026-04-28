import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Transfer, CreateTransferRequest, PaginatedResponse } from '../models/models';

@Injectable({ providedIn: 'root' })
export class TransferService {
  private http = inject(HttpClient);

  list(page = 1, limit = 20): Observable<PaginatedResponse<Transfer>> {
    const params = new HttpParams().set('page', page).set('limit', limit);
    return this.http.get<PaginatedResponse<Transfer>>('/api/transfers', { params });
  }

  create(data: CreateTransferRequest): Observable<Transfer> {
    return this.http.post<Transfer>('/api/transfers', data);
  }

  update(
    id: string,
    data: { name: string; date: string; tag_ids: string[] },
  ): Observable<Transfer> {
    return this.http.put<Transfer>(`/api/transfers/${id}`, data);
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`/api/transfers/${id}`);
  }
}
