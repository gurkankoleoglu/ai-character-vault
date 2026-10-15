export const getNotes = () => {
  const notes = localStorage.getItem('notes');
  return notes ? JSON.parse(notes) : [];
};

export const saveNotes = (notes) => {
  localStorage.setItem('notes', JSON.stringify(notes));
};

export const addNote = (note) => {
  const notes = getNotes();
  const newNote = {
    id: Date.now(),
    ...note,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };
  notes.push(newNote);
  saveNotes(notes);
  return newNote;
};

export const updateNote = (id, updatedNote) => {
  const notes = getNotes();
  const index = notes.findIndex((note) => note.id === id);
  if (index !== -1) {
    notes[index] = {
      ...notes[index],
      ...updatedNote,
      updatedAt: new Date().toISOString(),
    };
    saveNotes(notes);
    return notes[index];
  }
  return null;
};

export const deleteNote = (id) => {
  const notes = getNotes();
  const filteredNotes = notes.filter((note) => note.id !== id);
  saveNotes(filteredNotes);
  return filteredNotes;
};
