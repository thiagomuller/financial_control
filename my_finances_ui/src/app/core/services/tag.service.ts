import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { Tag } from '../models/models';

@Injectable({ providedIn: 'root' })
export class TagService {
  private http = inject(HttpClient);

  list(): Observable<Tag[]> {
    return this.http.get<Tag[]>('/api/tags');
  }

  create(name: string, color?: string): Observable<Tag> {
    return this.http.post<Tag>('/api/tags', { name, color: color ?? null });
  }

  update(id: string, name: string, color?: string): Observable<Tag> {
    return this.http.put<Tag>(`/api/tags/${id}`, { name, color: color ?? null });
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`/api/tags/${id}`);
  }
}
