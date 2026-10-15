import { getCategoryColor, getPriorityColor } from '../utils/aiAnalyzer';

export default function NoteCard({ note, onEdit, onDelete }) {
  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('tr-TR', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="bg-white rounded-2xl shadow-soft p-6 hover:shadow-soft-lg transition-all">
      <div className="flex justify-between items-start mb-3">
        <h3 className="text-xl font-bold text-background-900">{note.title}</h3>
        <div className="flex gap-2">
          <button
            onClick={() => onEdit(note)}
            className="text-primary-500 hover:text-primary-700 transition-colors"
            title="Düzenle"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
            </svg>
          </button>
          <button
            onClick={() => onDelete(note.id)}
            className="text-secondary-500 hover:text-secondary-700 transition-colors"
            title="Sil"
          >
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
      
      <p className="text-background-600 mb-4 line-clamp-3">{note.content}</p>
      
      <div className="flex flex-wrap gap-2 mb-3">
        <span className={`px-3 py-1 rounded-full text-sm font-medium ${getCategoryColor(note.category)}`}>
          {note.category}
        </span>
        <span className={`px-3 py-1 rounded-full text-sm font-medium ${getPriorityColor(note.priority)}`}>
          {note.priority}
        </span>
      </div>
      
      <p className="text-sm text-background-400">
        {formatDate(note.updatedAt)}
      </p>
    </div>
  );
}
